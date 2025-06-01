cmake_minimum_required(VERSION 3.10)
project(opus_audio_demo CXX)

# Set C++ standard
set(CMAKE_CXX_STANDARD 11)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Add executable
add_executable(opus_audio_demo main.cpp)

# Include directories
target_include_directories(opus_audio_demo PRIVATE
    ${CMAKE_SOURCE_DIR}/../../ogg/include/
    ${CMAKE_SOURCE_DIR}/../../opus/include/
)

# Link libraries
if(WIN32)
    # Windows specific settings
    target_link_libraries(opus_audio_demo PRIVATE
        ${CMAKE_SOURCE_DIR}/../../ogg/ogg.lib
        ${CMAKE_SOURCE_DIR}/../../opus/opus.lib
    )
else()
    # Unix-like systems
    find_package(PkgConfig REQUIRED)
    pkg_check_modules(OGG REQUIRED ogg)
    pkg_check_modules(OPUS REQUIRED opus)
    
    target_include_directories(opus_audio_demo PRIVATE
        ${OGG_INCLUDE_DIRS}
        ${OPUS_INCLUDE_DIRS}
    )
    
    target_link_libraries(opus_audio_demo PRIVATE
        ${OGG_LIBRARIES}
        ${OPUS_LIBRARIES}
        m  # for math library
    )
endif()

# Set compiler flags
if(MSVC)
    target_compile_options(opus_audio_demo PRIVATE /W4)
else()
    target_compile_options(opus_audio_demo PRIVATE -Wall -Wextra)
endif() 